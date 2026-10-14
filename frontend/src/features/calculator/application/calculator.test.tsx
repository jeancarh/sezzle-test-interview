import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { CalculationError, type CalculatorApi } from "../domain/calculatorApi";
import { CalculatorApiProvider, useCalculator } from "./calculator";

function setup(api: CalculatorApi) {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>
      <CalculatorApiProvider value={api}>{children}</CalculatorApiProvider>
    </QueryClientProvider>
  );
  return renderHook(() => useCalculator(), { wrapper });
}

describe("useCalculator", () => {
  it("sends the pending calculation to the port and shows the result", async () => {
    const calculate = vi.fn().mockResolvedValue("5");
    const { result } = setup({ calculate });

    act(() => {
      result.current.press({ kind: "digit", digit: "2" });
      result.current.press({ kind: "operator", operation: "add" });
      result.current.press({ kind: "digit", digit: "3" });
      result.current.press({ kind: "equals" });
    });
    expect(result.current.busy).toBe(true);

    await waitFor(() => expect(result.current.state.display).toBe("5"));
    expect(result.current.busy).toBe(false);
    expect(calculate).toHaveBeenCalledWith({ operation: "add", operands: ["2", "3"] });
  });

  it("puts the port's error message on screen", async () => {
    const calculate = vi.fn().mockRejectedValue(new CalculationError("division_by_zero", "Can't divide by zero"));
    const { result } = setup({ calculate });

    act(() => {
      result.current.press({ kind: "digit", digit: "1" });
      result.current.press({ kind: "operator", operation: "divide" });
      result.current.press({ kind: "digit", digit: "0" });
      result.current.press({ kind: "equals" });
    });

    await waitFor(() => expect(result.current.state.error).toBe("Can't divide by zero"));
  });

  it("uses a generic message when the port throws something unexpected", async () => {
    const calculate = vi.fn().mockRejectedValue("boom");
    const { result } = setup({ calculate });

    act(() => {
      result.current.press({ kind: "digit", digit: "9" });
      result.current.press({ kind: "sqrt" });
    });

    await waitFor(() => expect(result.current.state.error).toBe("Something went wrong, please try again"));
  });

  it("recalls a value from the tape", () => {
    const { result } = setup({ calculate: vi.fn() });
    act(() => result.current.recall("42"));
    expect(result.current.state.display).toBe("42");
  });

  it("fails loudly when the port is not provided", () => {
    const client = new QueryClient();
    const wrapper = ({ children }: { children: ReactNode }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>;
    vi.spyOn(console, "error").mockImplementation(() => undefined);
    expect(() => renderHook(() => useCalculator(), { wrapper })).toThrow(/CalculatorApi is not provided/);
  });
});
