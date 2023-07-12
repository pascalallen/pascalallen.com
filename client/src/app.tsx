import React, { ReactElement } from 'react';
import { createRoot } from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import routes from '@routes/routes';
import { storesInstance, StoresProvider } from '@stores/Stores';
import '@assets/scss/app.scss';

const container: HTMLElement | null = document.getElementById('root');
if (container === null) {
  throw new Error('No matching element found with ID: root');
}

const App = (): ReactElement => {
  return (
    <React.StrictMode>
      <StoresProvider value={storesInstance}>
        <RouterProvider router={createBrowserRouter(routes)} />
      </StoresProvider>
    </React.StrictMode>
  );
};

const root = createRoot(container);
root.render(<App />);
