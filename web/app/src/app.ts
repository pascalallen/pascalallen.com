import './styles/app.scss';
import env, { EnvKey } from './env';

const status = document.getElementById('status');

if (status !== null && env(EnvKey.APP_ENV) === 'local') {
  status.setAttribute('data-env', 'local');
}
