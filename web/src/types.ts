export type Ownership = "physical" | "digital" | "both" | "coveting";
export type ReadStatus = "unread" | "reading" | "read";

export type Named = {
  id: number;
  name: string;
};

export type Book = {
  id: number;
  title: string;
  subtitle?: string | null;
  isbn?: string | null;
  publishedYear?: number | null;
  description?: string | null;
  coverUrl?: string | null;
  seriesName?: string | null;
  seriesIndex?: number | null;
  ownedPhysical: boolean;
  ownedDigital: boolean;
  ownership: Ownership;
  readStatus: ReadStatus;
  wantRating?: number | null;
  qualityRating?: number | null;
  authors: Named[];
  piles?: Named[];
  collections?: Named[];
  genres?: Named[];
  createdAt: string;
  updatedAt: string;
};

export type SuggestedBook = Book & {
  score: number;
  reason: string;
};

export type Author = {
  id: number;
  name: string;
  bookCount: number;
  readCount: number;
  unreadWantSum: number;
  unreadOwned: number;
  unreadCoveting: number;
};

export type AuthorDetail = Author & {
  owned: Book[];
  coveting: Book[];
};

export type BookFilters = {
  ownership?: string;
  pileId?: string;
  collectionId?: string;
  genreId?: string;
  authorId?: string;
  readStatus?: string;
  minWant?: string;
  sort?: string;
  q?: string;
};
