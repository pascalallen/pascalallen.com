export enum EnvKey {
  APP_BASE_URL = 'APP_BASE_URL',
  APP_ENV = 'APP_ENV'
}

const raw = document.getElementById('script_config')?.getAttribute('value') ?? '';

const scriptConfig: Record<string, string> = raw === '' ? {} : JSON.parse(atob(raw));

const env = (key: EnvKey): string => scriptConfig[key] ?? '';

export default env;
