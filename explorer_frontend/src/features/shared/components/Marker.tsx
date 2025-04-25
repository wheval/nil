import { styled } from "styletron-react";

const Marker = styled("div", (props: any) => ({
  width: "8px",
  height: "8px",
  borderRadius: "50%",
  backgroundColor: props.$color,
}));

export { Marker };
