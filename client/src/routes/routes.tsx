import {RouteObject} from "react-router-dom";
import Path from "@domain/constants/Path";
import IndexPage from "@pages/IndexPage";
import LoginPage from "@pages/LoginPage";

const routes: RouteObject[] = [
    {
        path: Path.INDEX,
        element: <IndexPage/>
    },
    {
        path: Path.LOGIN,
        element: <LoginPage/>
    }
];

export default routes;