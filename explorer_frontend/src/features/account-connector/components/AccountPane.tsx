import { COLORS, LabelSmall } from "@nilfoundation/ui-kit";
import { ErrorBoundary } from "react-error-boundary";
import { useStyletron } from "styletron-react";
import { AccountContent } from "./AccountContent";
import { styles } from "./styles";

const AccountPane = () => {
  const [css] = useStyletron();

  return (
    <div className={css(styles.container)}>
      <ErrorBoundary
        fallback={
          <LabelSmall color={COLORS.red200}>There is a problem with the smart account</LabelSmall>
        }
      >
        <AccountContent />
      </ErrorBoundary>
    </div>
  );
};

export { AccountPane };
