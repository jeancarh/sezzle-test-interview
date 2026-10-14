import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { CalculatorApiProvider } from "../features/calculator/application/calculator";
import type { CalculatorApi } from "../features/calculator/domain/calculatorApi";
import { httpCalculatorApi } from "../features/calculator/infrastructure/httpCalculatorApi";
import { CalculatorPage } from "../features/calculator/ui/CalculatorPage";

export function createQueryClient() {
  return new QueryClient({ defaultOptions: { mutations: { retry: false } } });
}

export function Providers({ api, client, children }: { api: CalculatorApi; client: QueryClient; children: ReactNode }) {
  return (
    <QueryClientProvider client={client}>
      <CalculatorApiProvider value={api}>{children}</CalculatorApiProvider>
    </QueryClientProvider>
  );
}

const client = createQueryClient();

export function App() {
  return (
    <Providers api={httpCalculatorApi} client={client}>
      <CalculatorPage />
    </Providers>
  );
}
