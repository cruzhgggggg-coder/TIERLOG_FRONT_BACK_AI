declare module 'vue' {
  interface IntrinsicElements {
    'spline-viewer': {
      url?: string;
      loadingAnimType?: string;
      eventsTarget?: string;
      class?: string;
    } & Record<string, unknown>;
  }
}

export {};
