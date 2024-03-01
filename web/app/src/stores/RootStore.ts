import AuthStore from './AuthStore';

class RootStore {
  public authStore: AuthStore;

  constructor() {
    this.authStore = new AuthStore();
  }
}

export default RootStore;
