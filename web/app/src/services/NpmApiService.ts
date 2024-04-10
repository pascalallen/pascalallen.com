import axios, { AxiosInstance } from 'axios';
import { ApiResponse } from '@services/ApiService';

export type NpmPackage = {
  package: {
    name: string;
    version: string;
    description: string;
    links: {
      npm: string;
    };
  };
};

class NpmApiService {
  private readonly api: AxiosInstance;
  private readonly baseUrl: string = 'https://registry.npmjs.org';

  public constructor() {
    this.api = axios.create();
  }

  public async getAllPackages(): Promise<ApiResponse<NpmPackage[]>> {
    const response = await this.api(`${this.baseUrl}/-/v1/search?text=@pascalallen`);

    return {
      statusCode: response.status,
      body: {
        status: 'success',
        data: response.data.objects
      }
    };
  }
}

export default NpmApiService;
