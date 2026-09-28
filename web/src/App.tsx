import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { AddBook } from "./pages/AddBook";
import { AuthorDetailPage } from "./pages/AuthorDetail";
import { Authors } from "./pages/Authors";
import { GroupBrowse, GroupList } from "./pages/Groups";
import { ImportPage } from "./pages/Import";
import { Library } from "./pages/Library";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Library />} />
          <Route path="/authors" element={<Authors />} />
          <Route path="/authors/:id" element={<AuthorDetailPage />} />
          <Route path="/piles" element={<GroupList kind="piles" />} />
          <Route path="/piles/:id" element={<GroupBrowse kind="piles" />} />
          <Route path="/collections" element={<GroupList kind="collections" />} />
          <Route path="/collections/:id" element={<GroupBrowse kind="collections" />} />
          <Route path="/genres" element={<GroupList kind="genres" />} />
          <Route path="/genres/:id" element={<GroupBrowse kind="genres" />} />
          <Route path="/import" element={<ImportPage />} />
          <Route path="/add" element={<AddBook />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
