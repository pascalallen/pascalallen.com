import { Json } from '@domain/types/Json';

const scriptConfig = JSON.parse(atob(`${document.getElementById('script_config')?.getAttribute('value')}`));

export enum EnvKey {
  APP_BASE_URL = 'APP_BASE_URL',
  GITHUB_TOKEN = 'GITHUB_TOKEN'
}

const env = (key: EnvKey): Json => {
  return scriptConfig[key];
};

export default env;
