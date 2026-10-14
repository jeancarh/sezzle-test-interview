import { useMutation } from "@tanstack/react-query";
import { useCallback, useEffect, useReducer } from "react";
import { createPortContext } from "../../../shared/di/createPortContext";
import type { CalculationRequest, CalculatorApi } from "../domain/calculatorApi";
import { calculatorReducer, initialState, type Key } from "../domain/keypad";

export const [CalculatorApiProvider, useCalculatorApi] = createPortContext<CalculatorApi>("CalculatorApi");

export function useCalculator() {
  const api = useCalculatorApi();
  const [state, dispatch] = useReducer(calculatorReducer, initialState);
  const { mutateAsync } = useMutation({
    mutationFn: (request: CalculationRequest) => api.calculate(request),
  });

  const flight = state.inFlight;
  useEffect(() => {
    if (!flight) {
      return;
    }
    mutateAsync(flight.request).then(
      (value) => dispatch({ type: "resolved", id: flight.id, value }),
      (error: unknown) =>
        dispatch({
          type: "failed",
          id: flight.id,
          message: error instanceof Error ? error.message : "Something went wrong, please try again",
        }),
    );
  }, [flight, mutateAsync]);

  const press = useCallback((key: Key) => dispatch({ type: "key", key }), []);
  const recall = useCallback((value: string) => dispatch({ type: "recall", value }), []);

  return { state, press, recall, busy: flight !== null };
}
