import { Route, Routes } from "react-router-dom";
import { Layout } from "@/Layout";
import { ConvertPage } from "@/pages/ConvertPage";
import { ConversionsPage } from "@/pages/ConversionsPage";
import { HoldingsPage } from "@/pages/HoldingsPage";
import { DividendsPage } from "@/pages/DividendsPage";

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<ConvertPage />} />
        <Route path="conversions" element={<ConversionsPage />} />
        <Route path="holdings" element={<HoldingsPage />} />
        <Route path="dividends" element={<DividendsPage />} />
      </Route>
    </Routes>
  );
}
