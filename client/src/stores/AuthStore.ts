import {makeAutoObservable} from "mobx";
import {AuthData} from "../types/AuthData";

class AuthStore {
    private data?: AuthData;

    constructor() {
        makeAutoObservable(this);
    }

    public setData(data: AuthData): void {
        this.data = Object.freeze(data);
    }

    public clearData(): void {
        delete this.data;
    }

    public hasData(): boolean {
        return this.data !== undefined;
    }

    public getData(): AuthData | undefined {
        return this.data;
    }
}

export default AuthStore;