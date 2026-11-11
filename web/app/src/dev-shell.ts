import { installShell } from './bridge';
import { createFakeShell } from './fake-shell';

installShell(createFakeShell());
