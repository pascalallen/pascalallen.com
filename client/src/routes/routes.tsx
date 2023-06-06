import {RouteObject} from "react-router-dom";
import Path from "@domain/constants/Path";
import IndexPage from "@pages/IndexPage";
import LoginPage from "@pages/LoginPage";
import RouteElementWrapper from "@routes/middleware/RouteElementWrapper";
import RequiresAuthentication from "@routes/middleware/RequiresAuthentication";
import AuthPage from "@pages/AuthPage";

const routes: RouteObject[] = [
    {
        path: Path.INDEX,
        element: <IndexPage/>
    },
    {
        path: Path.LOGIN,
        element: <LoginPage/>
    },
    {
        path: '/auth',
        element: <RouteElementWrapper><RequiresAuthentication><AuthPage/></RequiresAuthentication></RouteElementWrapper>
    }
];

export default routes;