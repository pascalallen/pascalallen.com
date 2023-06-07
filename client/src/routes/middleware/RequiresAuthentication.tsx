import React, { ReactElement, ReactNode } from 'react';
import { observer } from 'mobx-react-lite';
import { useLocation } from 'react-router';
import { Navigate } from 'react-router-dom';
import Path from '@domain/constants/Path';
import useAuth from '@hooks/useAuth';

export type RequiresAuthenticationProps = {
  children: ReactNode;
};

const RequiresAuthentication = observer((props: RequiresAuthenticationProps): ReactElement => {
  const { children } = props;

  const authService = useAuth();
  const location = useLocation();

  if (!authService.isLoggedIn()) {
    // Redirect them to the /login page, but save the current location they
    // were trying to go to when they were redirected. This allows us to send
    // them along to that page after they log in, which is a nicer user
    // experience than dropping them off on the home page.
    return <Navigate to={Path.LOGIN} state={{ from: location }} replace />;
  }

  return <>{children}</>;
});

export default RequiresAuthentication;
