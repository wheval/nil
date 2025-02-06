import { styled } from "styletron-react";

export const Box = styled(
  "div",
  ({
    $padding,
    $align,
    $justify,
    $gap,
    $position,
    $transform,
    $top,
    $left,
    $width,
  }: {
    $padding?: string;
    $align?: string;
    $justify?: string;
    $gap?: string;
    $position?: "absolute" | "relative" | "static" | "fixed" | "sticky";
    $transform?: string;
    $top?: string;
    $left?: string;
    $width?: string;
  }) => ({
    display: "flex",
    flexDirection: "column",
    alignItems: $align || "stretch",
    justifyContent: $justify || "flex-start",
    padding: $padding || "0",
    gap: $gap || "0",
    width: $position === "absolute" ? $width || "auto" : "100%",
    boxSizing: "border-box",
    position: $position || "static",
    transform: $transform || "none",
    top: $top || "auto",
    left: $left || "auto",
    transition: "background-color 0.2s ease",
  }),
);
