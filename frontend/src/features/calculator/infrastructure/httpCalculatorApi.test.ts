import { afterEach, describe, expect, it, vi } from "vitest";
import { CalculationError } from "../domain/calculatorApi";
import { httpCalculatorApi } from "./httpCalculatorApi";

function respond(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": status < 400 ? "application/json" : "application/problem+json" },
    }),
  );
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("httpCalculatorApi", () => {
  it("posts the operands as strings and returns the result string", async () => {
    const fetchMock = respond(200, { operation: "add", operands: ["0.1", "0.2"], result: "0.3" });

    await expect(httpCalculatorApi.calculate({ operation: "add", operands: ["0.1", "0.2"] })).resolves.toBe("0.3");

    const [url, init] = fetchMock.mock.calls[0]!;
    expect(url).toBe("/api/v1/calculations");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body)).toEqual({ operation: "add", operands: ["0.1", "0.2"] });
  });

  it("turns problem codes into friendly messages", async () => {
    respond(422, { status: 422, code: "division_by_zero", detail: "can't divide by zero" });

    const error = await httpCalculatorApi.calculate({ operation: "divide", operands: ["1", "0"] }).catch((e) => e);

    expect(error).toBeInstanceOf(CalculationError);
    expect(error.code).toBe("division_by_zero");
    expect(error.message).toBe("Can't divide by zero");
  });

  it("falls back to a generic message for codes it doesn't know", async () => {
    respond(500, { status: 500, code: "internal_error", detail: "something went wrong on our side" });

    const error = await httpCalculatorApi.calculate({ operation: "add", operands: ["1", "2"] }).catch((e) => e);

    expect(error.code).toBe("internal_error");
    expect(error.message).toBe("Something went wrong, please try again");
  });

  it("explains network failures", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));

    const error = await httpCalculatorApi.calculate({ operation: "add", operands: ["1", "2"] }).catch((e) => e);

    expect(error.code).toBe("network_error");
    expect(error.message).toMatch(/backend/);
  });

  it("copes with error responses that aren't JSON", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("<html>bad gateway</html>", { status: 502, statusText: "Bad Gateway" })));

    const error = await httpCalculatorApi.calculate({ operation: "add", operands: ["1", "2"] }).catch((e) => e);

    expect(error.code).toBe("http_error");
  });
});
