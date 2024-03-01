import AuthService from '@services/AuthService';
import useStore from './useStore';

const useAuth = (): AuthService => {
  const authStore = useStore('authStore');

  return new AuthService(authStore);
};

export default useAuth;
