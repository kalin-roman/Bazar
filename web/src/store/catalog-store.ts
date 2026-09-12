import { create } from "zustand";
import { api, ApiCategory, ApiProduct } from "../lib/api";

type CatalogState = {
  categories: ApiCategory[];
  products: ApiProduct[];
  loading: boolean;
  error: string | null;
  fetch: () => Promise<void>;
};

// Replaces the static assets/categories.tsx + assets/products.ts
// arrays with a real fetch against the live backend. Kept as a
// zustand store (matching cart-store/orders-store's existing shape)
// rather than a per-screen fetch, so every screen shares one fetch
// and one cache instead of re-requesting the same data repeatedly.
export const useCatalogStore = create<CatalogState>((set) => ({
  categories: [],
  products: [],
  loading: false,
  error: null,
  fetch: async () => {
    set({ loading: true, error: null });
    try {
      const [categories, products] = await Promise.all([
        api.listCategories(),
        api.listProducts(),
      ]);
      set({ categories, products, loading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : "Failed to load the catalog",
        loading: false,
      });
    }
  },
}));
