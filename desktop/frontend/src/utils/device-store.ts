import {current} from '@reduxjs/toolkit';
import {
  DefinitionVersionMap,
  getTheme,
  KeyboardDefinitionIndex,
  KeyboardDictionary,
  keyboardDefinitionV2ToVIADefinitionV2,
  keyboardDefinitionV3ToVIADefinitionV3,
  ThemeDefinition,
} from '@the-via/reader';
import {TestKeyboardSoundsMode} from 'src/components/void/test-keyboard-sounds';
import {THEMES} from 'src/utils/themes';
import {Store} from '../shims/via-app-store';
import type {
  AuthorizedDevice,
  DefinitionIndex,
  Settings,
  VendorProductIdMap,
} from '../types/types';
import {getVendorProductId} from './hid-keyboards';
let deviceStore: Store;
const defaultStoreData = {
  definitionIndex: {
    generatedAt: -1,
    hash: '',
    version: '2.0.0',
    theme: getTheme(),
    accentColor: '#ad7070',
    supportedVendorProductIdMap: {},
  },
  definitions: {},
  settings: {
    showDesignTab: false,
    showConsoleTab: false,
    disableFastRemap: false,
    ShowSliderValuesMode: 'Slider Only' as const,
    renderMode: '2D' as const,
    themeMode: 'dark' as const,
    designDefinitionVersion: 'v3' as const,
    themeName: 'OLIVIA_DARK',
    hostKeyboardLayout: 'keymap_us',
    macroEditor: {
      smartOptimizeEnabled: true,
      recordDelaysEnabled: false,
      tapEnterAtEOMEnabled: false,
    },
    testKeyboardSoundsSettings: {
      isEnabled: true,
      volume: 100,
      waveform: 'sine' as const,
      mode: TestKeyboardSoundsMode.WickiHayden,
      transpose: 0,
    },
  },
};

function initDeviceStore() {
  deviceStore = new Store(defaultStoreData);
}

initDeviceStore();

// TODO: invalidate cache if we change cache structure

/** Retreives the latest definition index and invalidates the definition cache if a new one is found */
export async function syncStore(): Promise<DefinitionIndex> {
  const currentDefinitionIndex = deviceStore.get('definitionIndex');

  // TODO: fall back to cache if can't hit endpoint, notify user
  try {
    // Get hash file
    //    const hash = await (await fetch('/definitions/hash.json')).json();
    const hash = document.getElementById('definition_hash')?.dataset.hash || '';

    if (hash === currentDefinitionIndex.hash) {
      return currentDefinitionIndex;
    }
    // Get definition index file
    const response = await fetch('/definitions/supported_kbs.json', {
      cache: 'reload',
    });
    const json: KeyboardDefinitionIndex = await response.json();

    // TODO: maybe we should just export this shape from keyboards repo
    // v3 is a superset of v2 - if the def is avail in v2, it is also avail in v3
    const v2vpidMap = json.vendorProductIds.v2.reduce(
      (acc: VendorProductIdMap, id) => {
        acc[id] = acc[id] || {};
        acc[id].v2 = acc[id].v3 = true;
        return acc;
      },
      {},
    );

    const vpidMap = json.vendorProductIds.v3.reduce(
      (acc: VendorProductIdMap, def) => {
        acc[def] = acc[def] || {};
        acc[def].v3 = true;
        return acc;
      },
      v2vpidMap,
    );

    const newIndex = {
      ...json,
      hash,
      supportedVendorProductIdMap: vpidMap,
    };
    deviceStore.set('definitionIndex', newIndex);
    deviceStore.set('definitions', {});

    return newIndex;
  } catch (e) {
    console.warn(e);
  }

  return currentDefinitionIndex;
}

/**
 * Definitions shipped locally are authored in the raw KLE format
 * (`layouts.keymap`). The app expects the parsed `VIADefinitionV3` shape
 * (`layouts.keys`, `optionKeys`, ...). Upstream relies on an offline
 * build step for this; here we do the transform at load time so the
 * readable KLE files can be committed as-is.
 */
export const normalizeDefinition = <T>(
  definition: T,
  version: keyof DefinitionVersionMap,
): T => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const def = definition as any;
  if (!def || !def.layouts || def.layouts.keys || !def.layouts.keymap) {
    return definition;
  }
  try {
    return version === 'v3'
      ? (keyboardDefinitionV3ToVIADefinitionV3(def) as unknown as T)
      : (keyboardDefinitionV2ToVIADefinitionV2(def) as unknown as T);
  } catch (e) {
    console.error('Failed to parse keyboard definition', e);
    return definition;
  }
};

export const getMissingDefinition = async <
  K extends keyof DefinitionVersionMap,
>(
  device: AuthorizedDevice,
  version: K,
): Promise<[DefinitionVersionMap[K], K]> => {
  const vpid = getVendorProductId(device.vendorId, device.productId);
  const url = `/definitions/${version}/${vpid}.json`;
  const response = await fetch(url);
  const json: DefinitionVersionMap[K] = normalizeDefinition(
    await response.json(),
    version,
  );
  let definitions = deviceStore.get('definitions');
  const newDefinitions = {
    ...definitions,
    [vpid]: {
      ...definitions[vpid],
      [version]: json,
    },
  };

  try {
    deviceStore.set('definitions', newDefinitions);
  } catch (err) {
    // This is likely due to running out of space, so we clear it
    localStorage.clear();
    initDeviceStore();
    definitions = deviceStore.get('definitions');
    deviceStore.set('definitions', {
      ...definitions,
      [vpid]: {
        ...definitions[vpid],
        [version]: json,
      },
    });
  }
  return [json, version];
};

export const getSupportedIdsFromStore = (): VendorProductIdMap =>
  deviceStore.get('definitionIndex')?.supportedVendorProductIdMap;

export const getDefinitionsFromStore = (): KeyboardDictionary => {
  const definitions: KeyboardDictionary = deviceStore.get('definitions') || {};
  return Object.entries(definitions).reduce<KeyboardDictionary>(
    (acc, [vpid, versions]) => {
      acc[vpid] = {
        ...versions,
        v2: normalizeDefinition(versions?.v2, 'v2'),
        v3: normalizeDefinition(versions?.v3, 'v3'),
      };
      return acc;
    },
    {},
  );
};

export const getThemeFromStore = (): ThemeDefinition =>
  THEMES[getThemeNameFromStore() as keyof typeof THEMES] ||
  deviceStore.get('definitionIndex')?.theme;

export const getThemeModeFromStore = (): 'dark' | 'light' => {
  return deviceStore.get('settings')?.themeMode;
};

export const getShowSliderValuesModeFromStore = (): 'Slider & Show Value' | 'Slider & Input Field' | 'Slider Only' => {
  return deviceStore.get('settings')?.ShowSliderValuesMode;
};

export const getRenderModeFromStore = (): '3D' | '2D' => {
  return deviceStore.get('settings')?.renderMode;
};

export const getThemeNameFromStore = () => {
  return deviceStore.get('settings')?.themeName;
};

export const getSettings = (): Settings => deviceStore.get('settings');

export const setSettings = (settings: Settings) => {
  deviceStore.set('settings', current(settings));
};
