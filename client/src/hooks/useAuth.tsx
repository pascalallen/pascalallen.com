import useStore from '@hooks/useStore';
import AuthService from '@services/AuthService';

const useAuth = (): AuthService => {
  const authStore = useStore('authStore');

  return new AuthService(authStore);
};

export default useAuth;
