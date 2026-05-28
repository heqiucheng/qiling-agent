import { createContext } from "react";

import type { MockRole, MockUser } from "./auth-types";

export interface AuthContextValue {
  user: MockUser;
  switchRole: (role: MockRole) => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);
