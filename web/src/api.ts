import type {
  Author,
  AuthorDetail,
  Book,
  BookFilters,
  Named,
  SuggestedBook,
} from "./types";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* ignore */
    }
    throw new Error(message);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export function listBooks(filters: BookFilters = {}) {
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(filters)) {
    if (v) params.set(k, v);
  }
  const q = params.toString();
  return request<Book[]>(`/api/books${q ? `?${q}` : ""}`);
}

export const getBook = (id: number) => request<Book>(`/api/books/${id}`);

export const createBook = (body: Record<string, unknown>) =>
  request<Book>("/api/books", { method: "POST", body: JSON.stringify(body) });

export const patchBook = (id: number, body: Record<string, unknown>) =>
  request<Book>(`/api/books/${id}`, { method: "PATCH", body: JSON.stringify(body) });

export const putBookPiles = (id: number, pileIds: number[]) =>
  request<Book>(`/api/books/${id}/piles`, {
    method: "PUT",
    body: JSON.stringify({ pileIds }),
  });

export const putBookCollections = (id: number, collectionIds: number[]) =>
  request<Book>(`/api/books/${id}/collections`, {
    method: "PUT",
    body: JSON.stringify({ collectionIds }),
  });

export const listPiles = () => request<Named[]>("/api/piles");
export const createPile = (name: string) =>
  request<Named>("/api/piles", { method: "POST", body: JSON.stringify({ name }) });
export const getPile = (id: number) => request<Named>(`/api/piles/${id}`);

export const listCollections = () => request<Named[]>("/api/collections");
export const createCollection = (name: string) =>
  request<Named>("/api/collections", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
export const getCollection = (id: number) => request<Named>(`/api/collections/${id}`);

export const listGenres = () => request<Named[]>("/api/genres");
export const getGenre = (id: number) => request<Named>(`/api/genres/${id}`);

export const listAuthors = () => request<Author[]>("/api/authors");
export const getAuthor = (id: number) => request<AuthorDetail>(`/api/authors/${id}`);

export const previewCalibre = (libraryPath: string) =>
  request<{ books: number }>("/api/import/calibre/preview", {
    method: "POST",
    body: JSON.stringify({ libraryPath }),
  });

export const importCalibre = (libraryPath: string) =>
  request<{ imported: number; updated: number; covers: number; books: number }>(
    "/api/import/calibre",
    { method: "POST", body: JSON.stringify({ libraryPath }) },
  );

export const suggestNext = (limit = 12) =>
  request<SuggestedBook[]>(`/api/suggest/next?limit=${limit}`);

export const suggestAuthors = (limit = 12) =>
  request<Author[]>(`/api/suggest/authors?limit=${limit}`);
