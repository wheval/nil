import { COLORS } from "@nilfoundation/ui-kit";
import { ParagraphSmall } from "baseui/typography";
import type { FC } from "react";
import { useStyletron } from "styletron-react";
import arrowCircleUp from "../../assets/arrow-circle-up.svg";

export const SmartAccountNotConnectedWarning: FC = () => {
  const [css] = useStyletron();

  return (
    <div
      className={css({
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        alignItems: "center",
        paddingLeft: "40px",
        paddingRight: "40px",
        paddingTop: "24px",
        paddingBottom: "24px",
        gap: "24px",
      })}
    >
      <img
        src={arrowCircleUp}
        alt="Arrow pointing up to the account connector panel"
        className={css({
          width: "56px",
          height: "56px",
        })}
      />
      <ParagraphSmall
        className={css({
          textAlign: "center",
        })}
        color={COLORS.gray200}
      >
        Connect a Smart Account above to deploy contracts
      </ParagraphSmall>
    </div>
  );
};
