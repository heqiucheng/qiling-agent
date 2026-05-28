import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { IntentBadge } from "./IntentBadge";

describe("IntentBadge", () => {
  it("renders the high-intent label", () => {
    render(<IntentBadge level="high" />);
    expect(screen.getByText("高意向")).toBeInTheDocument();
  });
});