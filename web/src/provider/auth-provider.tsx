import type { Session, User } from "@supabase/supabase-js";
import {
  createContext,
  PropsWithChildren,
  useContext,
  useEffect,
  useState,
} from "react";
import { supabase } from "../lib/supabase";

type AuthResult = { error: string | null };

type AuthData = {
  session: Session | null;
  mounting: boolean;
  user: User | null;
  signIn: (email: string, password: string) => Promise<AuthResult>;
  signUp: (email: string, password: string) => Promise<AuthResult>;
  signOut: () => Promise<void>;
};

// Real Supabase auth — was a local mock (fabricated a user from
// whatever was typed, persisted to AsyncStorage under its own key)
// while the Go + Postgres backend didn't exist yet. Now that it does,
// this delegates to the real supabase-js client, which already
// handles its own session persistence/encryption (see supabase.ts's
// LargeSecureStore) and token refresh — nothing extra needed here.
const AuthContext = createContext<AuthData>({
  session: null,
  mounting: false,
  user: null,
  signIn: async () => ({ error: "Not implemented" }),
  signUp: async () => ({ error: "Not implemented" }),
  signOut: async () => {},
});

export default function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<Session | null>(null);
  const [mounting, setMounting] = useState(true);

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session);
      setMounting(false);
    });

    // Keeps session in sync with sign-in/sign-out/token-refresh events
    // that happen outside a direct call from this component (e.g. a
    // refreshed token, or a session restored from storage on cold
    // start finishing after the initial getSession() call above).
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(session);
    });

    return () => subscription.unsubscribe();
  }, []);

  const signIn = async (email: string, password: string): Promise<AuthResult> => {
    const { error } = await supabase.auth.signInWithPassword({ email, password });
    return { error: error?.message ?? null };
  };

  const signUp = async (email: string, password: string): Promise<AuthResult> => {
    const { error } = await supabase.auth.signUp({ email, password });
    return { error: error?.message ?? null };
  };

  const signOut = async () => {
    await supabase.auth.signOut();
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
