import { supabase } from "./supabase";

// Hardcoded, same as supabase.ts's own URL/anon key — this project's
// existing convention, rather than introducing an env-var mechanism
// the frontend doesn't otherwise use.
const API_BASE_URL = "https://bazar-xxjl.onrender.com";

// The Go backend's JSON responses use Go's exported field names
// as-is (no json struct tags anywhere yet), so these types are
// deliberately capitalized to match what's actually returned today,
// not conventional camelCase.
export type ApiCategory = {
  ID: number;
  Name: string;
  ImageURL: string;
  Slug: string;
};

export type ApiProduct = {
  ID: number;
  CategoryID: number;
  Title: string;
  Slug: string;
  ImagesURL: string[] | null;
  PriceCents: number;
  HeroImageURL: string;
  MaxQuantity: number;
};

export type ApiOrderItem = {
  ProductID: number;
  PriceCents: number;
  Quantity: number;
};

export type ApiOrder = {
  ID: number;
  UserID: string;
  Status: string;
  CreatedAt: string; // ISO 8601
  Items: ApiOrderItem[];
};

class ApiError extends Error {
  constructor(public status: number, path: string) {
    super(`API request to ${path} failed with status ${status}`);
  }
}

// Every route requires a valid Supabase-issued token — attach the
// current session's access_token as a Bearer header on every request.
// A caller with no session (shouldn't happen behind the app's own
// auth gate, but not assumed) sends no Authorization header at all,
// which the backend correctly rejects with 401.
async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (session?.access_token) {
    headers.set("Authorization", `Bearer ${session.access_token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });
  if (!response.ok) {
    throw new ApiError(response.status, path);
  }
  if (response.status === 204) {
    return undefined as T;
  }

  const body = await response.json();
  // A GET on an empty table encodes as JSON null (a nil Go slice),
  // not []  — normalize here so every caller can treat a list
  // response as a plain array without re-deriving this each time.
  return (body ?? []) as T;
}

export const api = {
  listCategories: () => apiFetch<ApiCategory[]>("/categories"),
  listProducts: () => apiFetch<ApiProduct[]>("/products"),
  listOrders: () => apiFetch<ApiOrder[]>("/orders"),
  getOrder: (id: number) => apiFetch<ApiOrder>(`/orders/${id}`),
  createOrder: (items: { product_id: number; quantity: number }[]) =>
    apiFetch<ApiOrder>("/orders", {
      method: "POST",
      body: JSON.stringify({ items }),
    }),
};
