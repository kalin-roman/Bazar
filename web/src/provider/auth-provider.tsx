import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  createContext,
  PropsWithChildren,
  useContext,
  useEffect,
  useState,
} from "react";

export type MockUser = {
  id: string;
  email: string;
  name: string;
};

export type MockSession = {
  user: MockUser;
};

type AuthResult = { error: string | null };

type AuthData = {
  session: MockSession | null;
  mounting: boolean;
  user: MockUser | null;
  signIn: (email: string, password: string) => Promise<AuthResult>;
  signUp: (email: string, password: string) => Promise<AuthResult>;
  signOut: () => Promise<void>;
};

const SESSION_STORAGE_KEY = "bazar-mock-session";

// There is no real backend yet (the Go + PostgreSQL API this will eventually
// call hasn't been built). Auth here is a local mock: it fabricates a user
// from whatever email/password the (already zod-validated) form submits, and
// persists the "session" to AsyncStorage so it survives app reloads. When the
// real API exists, only the bodies of signIn/signUp/signOut below need to
// change to real network calls — everything else (useAuth() shape, the auth
// gate, list-header.tsx) stays the same.
const AuthContext = createContext<AuthData>({
  session: null,
  mounting: false,
  user: null,
  signIn: async () => ({ error: "Not implemented" }),
  signUp: async () => ({ error: "Not implemented" }),
  signOut: async () => {},
});

export default function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<MockSession | null>(null);
  const [mounting, setMounting] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const stored = await AsyncStorage.getItem(SESSION_STORAGE_KEY);
        if (stored) {
          setSession(JSON.parse(stored) as MockSession);
        }
      } finally {
        setMounting(false);
      }
    })();
  }, []);

  const persistSession = async (newSession: MockSession | null) => {
    setSession(newSession);
    if (newSession) {
      await AsyncStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(newSession));
    } else {
      await AsyncStorage.removeItem(SESSION_STORAGE_KEY);
    }
  };

  const signIn = async (email: string, password: string): Promise<AuthResult> => {
    const user: MockUser = { id: email, email, name: email.split("@")[0] };
    await persistSession({ user });
    return { error: null };
  };

  const signUp = async (email: string, password: string): Promise<AuthResult> => {
    const user: MockUser = { id: email, email, name: email.split("@")[0] };
    await persistSession({ user });
    return { error: null };
  };

  const signOut = async () => {
    await persistSession(null);
  };

  return (
    <AuthContext.Provider
      value={{
        session,
        mounting,
        user: session?.user ?? null,
        signIn,
        signUp,
        signOut,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
