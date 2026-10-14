import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { CalculationError, type CalculationRequest, type CalculatorApi } from "../features/calculator/domain/calculatorApi";
import { CalculatorPage } from "../features/calculator/ui/CalculatorPage";
import { createQueryClient, Providers } from "./App";

const fakeMath: Record<string, (a: number, b: number) => number> = {
  add: (a, b) => a + b,
  subtract: (a, b) => a - b,
  multiply: (a, b) => a * b,
  divide: (a, b) => a / b,
  power: (a, b) => a ** b,
  percentage: (a, b) => (a * b) / 100,
};

function fakeApi(): CalculatorApi & { calculate: ReturnType<typeof vi.fn> } {
  return {
    calculate: vi.fn(async ({ operation, operands }: CalculationRequest) => {
      const [a, b] = operands.map(Number) as [number, number];
      if (operation === "divide" && b === 0) {
        throw new CalculationError("division_by_zero", "Can't divide by zero");
      }
      if (operation === "sqrt") {
        return String(Math.sqrt(a));
      }
      return String(fakeMath[operation]!(a, b));
    }),
  };
}

function renderCalculator(api = fakeApi()) {
  render(
    <Providers api={api} client={createQueryClient()}>
      <CalculatorPage />
    </Providers>,
  );
  return api;
}

const result = () => screen.getByRole("status", { name: "Result" });

async function click(user: ReturnType<typeof userEvent.setup>, ...names: string[]) {
  for (const name of names) {
    await user.click(screen.getByRole("button", { name }));
  }
}

describe("calculator page", () => {
  it("calculates with the mouse", async () => {
    const user = userEvent.setup();
    const api = renderCalculator();

    await click(user, "1", "Decimal point", "5", "Add", "2", "Equals");

    expect(await screen.findByText("3.5", { selector: "output" })).toBeInTheDocument();
    expect(screen.getByLabelText("Expression")).toHaveTextContent("1.5 + 2 =");
    expect(api.calculate).toHaveBeenCalledWith({ operation: "add", operands: ["1.5", "2"] });
  });

  it("calculates with the keyboard only", async () => {
    const user = userEvent.setup();
    renderCalculator();

    await user.keyboard("12*3{Enter}");

    expect(await screen.findByText("36", { selector: "output" })).toBeInTheDocument();
  });

  it("chains operations", async () => {
    const user = userEvent.setup();
    renderCalculator();

    await user.keyboard("2+3*4=");

    expect(await screen.findByText("20", { selector: "output" })).toBeInTheDocument();
  });

  it("shows API errors as an alert and recovers on the next key", async () => {
    const user = userEvent.setup();
    renderCalculator();

    await user.keyboard("1/0{Enter}");

    expect(await screen.findByRole("alert")).toHaveTextContent("Can't divide by zero");
    expect(result()).toHaveAttribute("aria-invalid", "true");

    await user.keyboard("7");
    expect(screen.getByRole("alert")).toHaveTextContent("");
    expect(result()).toHaveTextContent("7");
  });

  it("highlights the operator waiting for its second number", async () => {
    const user = userEvent.setup();
    renderCalculator();

    await click(user, "9", "Multiply");
    expect(screen.getByRole("button", { name: "Multiply" })).toHaveAttribute("aria-pressed", "true");

    await click(user, "3");
    expect(screen.getByRole("button", { name: "Multiply" })).toHaveAttribute("aria-pressed", "false");
  });

  it("edits the number being typed", async () => {
    const user = userEvent.setup();
    renderCalculator();

    await user.keyboard("123{Backspace}n");
    expect(result()).toHaveTextContent("-12");

    await user.keyboard("{Escape}");
    expect(result()).toHaveTextContent("0");
  });

  it("does square roots and percentages", async () => {
    const user = userEvent.setup();
    const api = renderCalculator();

    await click(user, "1", "6", "Square root");
    expect(await screen.findByText("4", { selector: "output" })).toBeInTheDocument();

    await user.keyboard("{Escape}15%200=");
    expect(await screen.findByText("30", { selector: "output" })).toBeInTheDocument();
    expect(api.calculate).toHaveBeenLastCalledWith({ operation: "percentage", operands: ["15", "200"] });
  });

  it("keeps recent results and reuses one on click", async () => {
    const user = userEvent.setup();
    const api = renderCalculator();

    await user.keyboard("6*7=");
    const tape = screen.getByRole("region", { name: "Recent" });
    const entry = await within(tape).findByRole("button", { name: "Use 42 (6 × 7)" });

    await click(user, "Add");
    await user.click(entry);
    await click(user, "Equals");

    expect(await screen.findByText("84", { selector: "output" })).toBeInTheDocument();
    expect(api.calculate).toHaveBeenLastCalledWith({ operation: "add", operands: ["42", "42"] });
  });

  it("doesn't double press the focused key when Enter is typed", async () => {
    const user = userEvent.setup();
    const api = renderCalculator();

    await click(user, "2", "Add", "3");
    expect(screen.getByRole("button", { name: "3" })).toHaveFocus();
    await user.keyboard("{Enter}");

    expect(await screen.findByText("5", { selector: "output" })).toBeInTheDocument();
    expect(api.calculate).toHaveBeenCalledTimes(1);
    expect(api.calculate).toHaveBeenCalledWith({ operation: "add", operands: ["2", "3"] });
  });

  it("treats a double click on equals as repeating the last operation", async () => {
    const user = userEvent.setup();
    const api = renderCalculator();

    await user.keyboard("2+3");
    await user.dblClick(screen.getByRole("button", { name: "Equals" }));

    expect(await screen.findByText("8", { selector: "output" })).toBeInTheDocument();
    expect(api.calculate).toHaveBeenNthCalledWith(2, { operation: "add", operands: ["5", "3"] });
  });

  it("ignores shortcuts with modifier keys", async () => {
    const user = userEvent.setup();
    renderCalculator();

    await user.keyboard("{Control>}5{/Control}");
    expect(result()).toHaveTextContent("0");
  });

  it("gives every key an accessible name", () => {
    renderCalculator();
    const keypad = screen.getByRole("group", { name: "Keypad" });
    const buttons = within(keypad).getAllByRole("button");
    expect(buttons).toHaveLength(22);
    for (const button of buttons) {
      expect(button).toHaveAccessibleName();
    }
  });
});
