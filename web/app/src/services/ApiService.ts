import axios, { AxiosError, AxiosInstance, AxiosRequestConfig } from 'axios';
import { StatusCodes } from 'http-status-codes';
import { jwtDecode } from 'jwt-decode';
import moment from 'moment';
import { listToMap } from '@utilities/collections';
import { Json } from '@domain/types/Json';
import AuthStore from '@stores/AuthStore';
import { AuthenticatedResponsePayload } from './AuthService';

export type SuccessResponseBody<T> = {
  status: 'success';
  data?: T;
};

export type ApiResponse<T> = {
  statusCode: number;
  body: SuccessResponseBody<T>;
};

export type FailResponseBody = {
  status: 'fail';
  data: { [field: string]: string };
};

export type FailApiResponse = {
  statusCode: number;
  body: FailResponseBody;
};

export type ErrorResponseBody = {
  status: 'error';
  message: string;
  code?: string;
  data?: { [field: string]: string };
};

export type ErrorApiResponse = {
  statusCode: number;
  body: ErrorResponseBody;
};

type RequestBody = Json | FormData;

type AppError = {
  message: string;
};

type Options = {
  auth: boolean;
  authStore?: AuthStore;
  contentType?: string;
};

const addAuthorizationHeader = (api: AxiosInstance, authStore?: AuthStore): void => {
  const authData = authStore?.getData();
  const accessToken = authData?.access_token;

  if (accessToken) {
    api.interceptors.request.use(config => {
      if (config.headers) {
        config.headers.Authorization = `Bearer ${accessToken}`;

        return config;
      }

      return config;
    });
  }
};

const addContentType = (api: AxiosInstance, contentType: string): void => {
  api.interceptors.request.use(config => {
    if (config.headers) {
      config.headers['Content-Type'] = contentType;
    }

    return config;
  });
};

export class ApiService {
  private readonly api: AxiosInstance;
  private readonly authStore?: AuthStore;
  private readonly contentType?: string;
  private readonly readConfig: AxiosRequestConfig = { headers: { Accept: 'application/json' } };
  private readonly writeConfig: AxiosRequestConfig = {
    headers: { 'Accept': 'application/json', 'Content-Type': 'application/json' }
  };

  public constructor(options: Options) {
    this.api = axios.create();
    this.authStore = options.authStore;
    this.contentType = options.contentType;

    if (this.contentType) {
      addContentType(this.api, this.contentType);
    }

    if (options.auth) {
      if (!this.authStore) {
        throw makeAppError('Must pass auth store when auth needed');
      }

      if (this.isTokenExpired()) {
        this.refreshToken().then(() => addAuthorizationHeader(this.api, this.authStore));

        return;
      }

      addAuthorizationHeader(this.api, this.authStore);
    }
  }

  public async get<T>(url: string): Promise<ApiResponse<T>> {
    const axiosResponse = await this.api.get(url, this.readConfig);

    return {
      statusCode: axiosResponse.status,
      body: axiosResponse.data
    };
  }

  public async post<T>(url: string, body: RequestBody = {}): Promise<ApiResponse<T>> {
    const axiosResponse = await this.api.post(url, body, this.writeConfig);

    return {
      statusCode: axiosResponse.status,
      body: axiosResponse.data
    };
  }

  public async put<T>(url: string, body: RequestBody = {}): Promise<ApiResponse<T>> {
    const axiosResponse = await this.api.put(url, body, this.writeConfig);

    return {
      statusCode: axiosResponse.status,
      body: axiosResponse.data
    };
  }

  public async patch<T>(url: string, body: RequestBody = {}): Promise<ApiResponse<T>> {
    const axiosResponse = await this.api.patch(url, body, this.writeConfig);

    return {
      statusCode: axiosResponse.status,
      body: axiosResponse.data
    };
  }

  public async delete<T>(url: string): Promise<ApiResponse<T>> {
    const axiosResponse = await this.api.delete(url, this.writeConfig);

    return {
      statusCode: axiosResponse.status,
      body: axiosResponse.data
    };
  }

  private async refreshToken(): Promise<void> {
    try {
      if (this.authStore) {
        const response = await this.patch<AuthenticatedResponsePayload>('/api/v1/auth/refresh', {
          access_token: this.authStore.getData()?.access_token ?? '',
          refresh_token: this.authStore.getData()?.refresh_token ?? ''
        });

        this.authStore.setData({
          access_token: response.body.data?.access_token ?? '',
          refresh_token: response.body.data?.refresh_token ?? '',
          access_token_claims: jwtDecode(response.body.data?.access_token ?? ''),
          client_iat: moment().unix(),
          roles: listToMap(response.body.data?.roles ?? []),
          permissions: listToMap(response.body.data?.permissions ?? [])
        });
      }
    } catch (err) {
      this.authStore?.clearData();
    }
  }

  private isTokenExpired(): boolean {
    const authData = this.authStore?.getData();

    if (!authData) {
      return true;
    }

    try {
      const clientIssuedAtTime = authData?.client_iat ?? 0;
      const serverIssuedAtTime = authData?.access_token_claims?.iat ?? 0;
      const serverExpTimestamp = authData?.access_token_claims?.exp ?? 0;
      const clientNowTimestamp = moment().unix();
      const clientServerDifference = clientIssuedAtTime - serverIssuedAtTime || 0;
      const calculatedServerNowTimestamp = clientNowTimestamp - clientServerDifference;
      const serverTokenExpTime = moment.unix(serverExpTimestamp || 0).subtract(5, 'minutes');
      const serverNowTimestamp = moment.unix(calculatedServerNowTimestamp);
      return serverNowTimestamp.isSameOrAfter(serverTokenExpTime);
    } catch (err) {
      return true;
    }
  }
}

export const makeAppError = (errorMessage?: string): AppError => {
  return { message: errorMessage || 'An unexpected error occurred' };
};

export const makeApiErrorResponse = (error: AxiosError<ErrorResponseBody>): ErrorApiResponse => {
  return {
    statusCode: error.response?.status || StatusCodes.INTERNAL_SERVER_ERROR,
    body: {
      status: error.response?.data?.status || 'error',
      message: error.response?.data?.message || 'API ERROR',
      code: error.code
    }
  };
};

export const makeApiFailResponse = (error: AxiosError<FailResponseBody>): FailApiResponse => {
  return {
    statusCode: error.response?.status || StatusCodes.INTERNAL_SERVER_ERROR,
    body: {
      status: error.response?.data?.status || 'fail',
      data: error.response?.data.data || {}
    }
  };
};

export const makeApiService = (options: Options): ApiService => {
  return new ApiService(options);
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const isApiFail = (error: any | unknown): error is FailApiResponse => {
  return (<FailApiResponse>error).body !== undefined && (<FailApiResponse>error).body.status === 'fail';
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const isApiError = (error: any | unknown): error is ErrorApiResponse => {
  return (<ErrorApiResponse>error).body !== undefined && (<ErrorApiResponse>error).body.status === 'error';
};

export default ApiService;
