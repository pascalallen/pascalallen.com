import React from 'react';
import RootStore from './RootStore';

export const storesInstance = new RootStore();
export const StoresContext = React.createContext(storesInstance);
export const StoresProvider = StoresContext.Provider;
