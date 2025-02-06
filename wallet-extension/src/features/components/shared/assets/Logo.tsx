import { styled } from "styletron-react";
import logoSrc from "../../../../../public/icons/logo/logo.svg";

// Define the styled component with inline typing
const StyledLogo = styled("img", ({ $size }: { $size: number }) => ({
  height: `${$size}px`,
  width: "auto",
}));

// Define the Logo component with props typing
export const Logo = ({ size = 60, alt = "Logo" }: { size?: number; alt?: string }) => {
  return <StyledLogo src={logoSrc} alt={alt} $size={size} draggable="false" />;
};
