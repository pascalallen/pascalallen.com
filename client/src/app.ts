import Path from "@domain/constants/Path";
import Index from "@pages/Index";
import Login from "@pages/Login";

const container: HTMLElement | null = document.getElementById('root');
if (container === null) {
    throw new Error('No matching element found with ID: root');
}

switch (location.pathname) {
    case Path.INDEX:
        Index().then(content => container.innerHTML = content);
        break;
    case Path.LOGIN:
        Login().then(content => container.innerHTML = content);
        break;
}
