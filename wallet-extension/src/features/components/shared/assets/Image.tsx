import { styled } from "styletron-react";

export const Image = styled(
  "img",
  ({
    $marginBottom,
    $height,
    $width,
  }: {
    $marginBottom?: string;
    $height?: string;
    $width?: string;
  }) => ({
    maxWidth: $width || "100%",
    height: $height || "auto",
    marginBottom: $marginBottom || "24px",
  }),
);
