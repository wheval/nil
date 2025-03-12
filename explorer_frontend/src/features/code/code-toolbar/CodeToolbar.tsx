import type { FC, ReactNode } from "react";
import { useStyletron } from "styletron-react";
import { BackRouterNavigationButton, useMobile } from "../../shared";
import { HyperlinkButton } from "./HyperlinkButton";
import { OpenProjectButton } from "./OpenProjectButton.tsx";
import { QuestionButton } from "./QuestionButton";

type CodeToolbarProps = {
  disabled: boolean;
  compilerVersionButton?: ReactNode;
};

export const CodeToolbar: FC<CodeToolbarProps> = ({ disabled, compilerVersionButton }) => {
  const [css] = useStyletron();
  const [isMobile] = useMobile();

  return (
    <div
      className={css({
        display: "flex",
        alignItems: "center",
        justifyContent: isMobile ? "flex-end" : "flex-start",
        gap: "8px",
        flexGrow: 1,
      })}
    >
      {!isMobile && (
        <BackRouterNavigationButton
          overrides={{
            Root: {
              style: {
                marginRight: "auto",
              },
            },
          }}
        />
      )}
      <QuestionButton />
      <HyperlinkButton disabled={disabled} />
      <OpenProjectButton disabled={disabled} />
      {compilerVersionButton}
    </div>
  );
};
