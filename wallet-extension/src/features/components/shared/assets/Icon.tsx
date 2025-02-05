import { COLORS } from "@nilfoundation/ui-kit";
import { styled } from "styletron-react";

const IconWrapper = styled(
  "span",
  ({
    $size,
    $round,
    $background,
    $hoverBackground,
    $pointer,
    $margin,
    $hasOnClick,
  }: {
    $size: number;
    $round: boolean;
    $background?: string;
    $hoverBackground?: string;
    $pointer?: boolean;
    $margin?: string;
    $hasOnClick?: () => void;
  }) => ({
    display: "inline-flex",
    justifyContent: "center",
    alignItems: "center",
    width: `${$size}px`,
    height: `${$size}px`,
    borderRadius: $round ? "50%" : "7px",
    backgroundColor: $background || COLORS.gray900,
    cursor: $pointer ? "pointer" : "default",
    transition: "background-color 0.2s ease",
    margin: $margin || "0",
    ":hover": {
      backgroundColor: $hoverBackground || "transparent",
    },
    ...($hasOnClick && { ":active": { transform: "scale(0.95)" } }),
  }),
);

const StyledIcon = styled("img", ({ $iconSize }: { $iconSize?: string }) => ({
  width: $iconSize || "60%",
  height: $iconSize || "60%",
}));

export const Icon = ({
  src,
  alt,
  size = 34,
  round = true,
  background,
  hoverBackground,
  iconSize,
  onClick,
  pointer,
  margin,
}: {
  src: {};
  alt: string;
  size?: number;
  round?: boolean;
  background?: string;
  hoverBackground?: string;
  iconSize?: string;
  onClick?: () => void;
  pointer?: boolean;
  margin?: string;
}) => {
  return (
    <IconWrapper
      $size={size}
      $round={round}
      $background={background}
      $hoverBackground={hoverBackground}
      $pointer={pointer}
      $hasOnClick={onClick}
      onClick={onClick}
      $margin={margin}
    >
      <StyledIcon src={src} alt={alt} $iconSize={iconSize} draggable="false" />
    </IconWrapper>
  );
};
