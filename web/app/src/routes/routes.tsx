import React from 'react';
import { RouteObject } from 'react-router-dom';
import Path from '../domain/constants/Path';
import RequiresAuthentication from './middleware/RequiresAuthentication';
import RouteElementWrapper from './middleware/RouteElementWrapper';
import IndexPage from '../pages/IndexPage';
import LoginPage from '../pages/LoginPage';
import TempPage from '../pages/TempPage';

const routes: RouteObject[] = [
  {
    path: Path.INDEX,
    element: <IndexPage />
  },
  {
    path: Path.LOGIN,
    element: <LoginPage />
  },
  {
    path: Path.TEMP,
    element: (
      <RouteElementWrapper>
        <RequiresAuthentication>
          <TempPage />
        </RequiresAuthentication>
      </RouteElementWrapper>
    )
  }
];

export default routes;
