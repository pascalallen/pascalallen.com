import React, { ReactElement, useEffect } from 'react';
import useStore from '@hooks/useStore';
import TempService from '@services/TempService';

const TempPage = (): ReactElement => {
  const authStore = useStore('authStore');

  useEffect(() => {
    const tempService = new TempService(authStore);
    tempService.temp().then(
      r => console.log(r),
      e => console.error(e)
    );
  }, [authStore]);

  return <>You&apos;re authenticated!</>;
};

export default TempPage;
