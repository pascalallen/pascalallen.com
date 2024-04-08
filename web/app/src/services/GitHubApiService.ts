import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { queryStringify, removeEmptyKeys } from '@utilities/collections';
import env, { EnvKey } from '@utilities/env';
import { ApiResponse } from '@services/ApiService';

export type GitHubRepositoryQueryParams = {
  type: 'all' | 'owner' | 'member';
  sort?: 'created' | 'updated' | 'pushed' | 'full_name';
  direction?: 'asc' | 'desc';
  per_page?: number;
  page?: number;
};

export type GitHubRepository = {
  name: string;
  html_url: string;
  updated_at: string;
  description: string;
};

class GitHubApiService {
  private readonly api: AxiosInstance;
  private readonly baseUrl: string = 'https://api.github.com';
  private readonly config: AxiosRequestConfig = {
    headers: {
      'Accept': 'application/vnd.github+json',
      'Authorization': `Bearer ${env(EnvKey.GITHUB_TOKEN)}`,
      'X-GitHub-Api-Version': '2022-11-28'
    }
  };

  public constructor() {
    this.api = axios.create();
  }

  public async getAllRepositories(params: GitHubRepositoryQueryParams): Promise<ApiResponse<GitHubRepository[]>> {
    const queryParams = queryStringify(removeEmptyKeys(params));
    const response = await this.api(`${this.baseUrl}/user/repos${queryParams}`, this.config);

    return {
      statusCode: response.status,
      body: {
        status: 'success',
        data: response.data
      }
    };
  }
}

export default GitHubApiService;
