import React, { ReactElement, ReactNode, useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import useAuth from '@hooks/useAuth';

type State = {
  refreshing: boolean;
  refreshAttempted: boolean;
};

const initialState: State = {
  refreshing: false,
  refreshAttempted: false
};

type Props = {
  children: ReactNode;
};

const RouteElementWrapper = observer((props: Props): ReactElement => {
  const { children } = props;

  const authService = useAuth();

  const [refreshing, setRefreshing] = useState(initialState.refreshing);
  const [refreshAttempted, setRefreshAttempted] = useState(initialState.refreshAttempted);

  useEffect(() => {
    if (!authService.isLoggedIn() && !refreshAttempted && !refreshing) {
      setRefreshing(true);
      authService
        .refresh()
        .then(() => {
          setRefreshAttempted(true);
          setRefreshing(false);
        })
        .catch(error => {
          console.error(error);
          setRefreshAttempted(true);
          setRefreshing(false);
        });
    }
  }, [authService, setRefreshing, setRefreshAttempted, refreshing, refreshAttempted]);

  if (!authService.isLoggedIn() && (!refreshAttempted || refreshing)) {
    return <>Loading...</>; // TODO: Return "loading" component
  }

  return <>{children}</>;
});

export default RouteElementWrapper;
