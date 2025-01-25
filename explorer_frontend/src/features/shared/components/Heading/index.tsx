import {
  HeadingLarge,
  HeadingXLarge,
  PRIMITIVE_COLORS,
  ParagraphMedium,
} from "@nilfoundation/ui-kit";
import type { BlockOverrides } from "baseui/block";
import { useStyletron } from "styletron-react";
import type { StylesObject } from "../..";
import { useMobile } from "../../hooks/useMobile";

type HeadingProps = {
  className?: string;
};

const subTextOverrides: BlockOverrides = {
  Block: {
    style: {
      color: PRIMITIVE_COLORS.gray300,
    },
  },
};

const s: StylesObject = {
  heading: {
    marginBottom: "4px",
  },
};

export const Heading = ({ className = "" }: HeadingProps) => {
  const [css] = useStyletron();

  const [isMobile] = useMobile();

  const TitleComponent = isMobile ? HeadingXLarge : HeadingLarge;

  return (
    <div
      className={`${css({
        flex: 0,
      })} ${className}`.trim()}
    >
      <TitleComponent className={css(s.heading)}>Secure Ethereum scaling</TitleComponent>
      <ParagraphMedium overrides={subTextOverrides}>
        =nil; is a zkRollup that scales Ethereum through zkSharding, enabling secure parallel
        execution of transactions across shards operating independently.
      </ParagraphMedium>
    </div>
  );
};
