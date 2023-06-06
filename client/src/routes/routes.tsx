import {RouteObject} from "react-router-dom";
import Path from "@domain/constants/Path";
import IndexPage from "@pages/IndexPage";

const routes: RouteObject[] = [
    {
        path: Path.INDEX,
        element: <IndexPage/>
    }
];

export default routes;