import React from 'react';
import { storesInstance, StoresContext } from '@stores/Stores';

const useStore = <T extends keyof typeof storesInstance>(store: T): (typeof storesInstance)[T] => {
  const storeContext = React.useContext(StoresContext)[store];

  if (!storeContext) {
    throw new Error('invalid store');
  }

  return storeContext;
};

export default useStore;
