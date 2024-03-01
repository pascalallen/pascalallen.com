import { makeAutoObservable } from 'mobx';
import { AuthData } from '@domain/types/AuthData';

class AuthStore {
  private data?: AuthData;

  constructor() {
    const serializedData = localStorage.getItem('auth_data');
    if (serializedData !== null) {
      this.data = { ...JSON.parse(serializedData) };
    }

    makeAutoObservable(this);
  }

  public setData(data: AuthData): void {
    this.data = Object.freeze(data);
    localStorage.setItem('auth_data', JSON.stringify(this.data));
  }

  public clearData(): void {
    this.data = undefined;
    localStorage.removeItem('auth_data');
  }

  public hasData(): boolean {
    return this.data !== undefined;
  }

  public getData(): AuthData | undefined {
    return this.data;
  }
}

export default AuthStore;
