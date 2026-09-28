import { describe, expect, it } from "vitest";
import { ownershipFromFlags, ownershipLabel } from "./ownership";

describe("ownershipFromFlags", () => {
  it("derives coveting when unowned", () => {
    expect(ownershipFromFlags(false, false)).toBe("coveting");
  });

  it("derives both when owned on two platforms", () => {
    expect(ownershipFromFlags(true, true)).toBe("both");
  });
});

describe("ownershipLabel", () => {
  it("labels coveting", () => {
    expect(ownershipLabel("coveting")).toBe("Coveting");
  });
});
