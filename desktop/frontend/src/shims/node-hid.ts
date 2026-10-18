import type {AuthorizedDevice, ConnectedDevice, WebVIADevice} from '../types/types';

const DEVICE_PATH = 'atom-he68-pro';
const DEVICE_INFO = {
  vendorId: 0x0c45,
  productId: 0x80cc,
  productName: 'ATOM HE68 PRO',
};

export const QMK_CONSOLE_FILTER = {
  usagePage: 0xff31,
  usage: 0x74,
};

export const isQMKConsoleDevice = () => false;

export const tryForgetDevice = (_device: ConnectedDevice | AuthorizedDevice) =>
  Promise.resolve();

declare global {
  interface Window {
    go: {
      main: {
        App: {
          Connected(): Promise<boolean>;
          HIDCommand(command: number, payload: number[]): Promise<number[]>;
        };
      };
    };
  }
}

const ExtendedHID = {
  _cache: {} as {[key: string]: WebVIADevice},
  requestDevice: async () => {
    return ExtendedHID.devices(true);
  },
  getFilteredDevices: async () => {
    return ExtendedHID.devices(false);
  },
  devices: async (_requestAuthorize = false) => {
    const connected = await window.go.main.App.Connected();
    if (!connected) {
      return [];
    }
    const device = {
      _device: {} as HIDDevice,
      __path: DEVICE_PATH,
      interface: 0x0001,
      vendorId: DEVICE_INFO.vendorId,
      productId: DEVICE_INFO.productId,
      path: DEVICE_PATH,
      productName: DEVICE_INFO.productName,
    };
    ExtendedHID._cache[DEVICE_PATH] = device as WebVIADevice;
    return [device as WebVIADevice];
  },
  HID: class HID {
    _hidDevice?: WebVIADevice;
    interface: number = 0x0001;
    vendorId: number = DEVICE_INFO.vendorId;
    productId: number = DEVICE_INFO.productId;
    productName: string = DEVICE_INFO.productName;
    path: string = DEVICE_PATH;
    openPromise: Promise<void> = Promise.resolve();
    private pendingResponse?: Uint8Array;

    constructor(path: string) {
      this._hidDevice = ExtendedHID._cache[path];
      if (!this._hidDevice) {
        throw new Error('Missing hid device in cache');
      }
      this.path = path;
    }

    async open() {
      return Promise.resolve();
    }

    setupListeners() {}

    addInputReportHandler() {
      return () => {};
    }

    read(fn: (err?: Error, data?: ArrayBuffer) => void) {
      fn(undefined, this.pendingResponse as any);
    }

    readP = () =>
      Promise.resolve(this.pendingResponse || new Uint8Array(0));

    async write(arr: number[]) {
      const command = arr[1];
      const payload = arr.slice(2);
      const resp = await window.go.main.App.HIDCommand(command, payload);
      this.pendingResponse = new Uint8Array(resp);
    }
  },
};

export const HID = ExtendedHID;
