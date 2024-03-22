import React, { ReactElement } from 'react';
import { createRoot } from 'react-dom/client';
import { Helmet, HelmetProvider } from 'react-helmet-async';
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
        <HelmetProvider>
          <Helmet>
            <title>Pascal Allen - pascalallen.com</title>
            <meta
              name="description"
              content="pascalallen.com is a containerized web application built with Kubernetes, Docker, RabbitMQ, Postgres, Go, React, TypeScript, Sass, Webpack, and WebAssembly. This ongoing project is designed, developed, deployed, and maintained by myself, Pascal Allen."
            />
            <meta name="robots" content="index, follow" />
            <link rel="canonical" href="https://pascalallen.com/" />
          </Helmet>
          <RouterProvider router={createBrowserRouter(routes)} />
        </HelmetProvider>
      </StoresProvider>
    </React.StrictMode>
  );
};

const root = createRoot(container);
root.render(<App />);
