import { COLORS, LabelMedium } from "@nilfoundation/ui-kit";
import type { FC } from "react";
import { useStyletron } from "styletron-react";

type TransactionEventLogProps = {
  message: string;
};

export const TransactionEventLog: FC<TransactionEventLogProps> = ({ message }) => {
  const [css] = useStyletron();

  return (
    <div
      className={css({
        display: "flex",
        gap: "8px",
        alignItems: "top",
      })}
    >
      <LabelMedium color={COLORS.gray400}>Event:</LabelMedium>
      <LabelMedium color={COLORS.gray50}>{message}</LabelMedium>
    </div>
  );
};
