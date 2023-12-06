import React, { ReactElement, useEffect } from 'react';
import { observer } from 'mobx-react-lite';
import { useNavigate } from 'react-router-dom';
import Path from '@domain/constants/Path';
import useStore from '@hooks/useStore';
import AuthService from '@services/AuthService';
import TempService from '@services/TempService';

const TempPage = observer((): ReactElement => {
  const authStore = useStore('authStore');
  const navigate = useNavigate();

  useEffect(() => {
    const tempService = new TempService(authStore);
    tempService.temp().then(
      r => console.log(r),
      e => console.error(e)
    );
  }, [authStore]);

  const handleLogout = async (): Promise<void> => {
    const authService = new AuthService(authStore);
    authService.logout().finally(() => navigate(Path.INDEX));
  };

  return (
    <>
      You&apos;re authenticated!
      <br />
      <button type="button" onClick={handleLogout}>
        Logout
      </button>
    </>
  );
});

export default TempPage;
