import AuthStore from '@stores/AuthStore';

class RootStore {
  public authStore: AuthStore;

  constructor() {
    this.authStore = new AuthStore();
  }
}

export default RootStore;
