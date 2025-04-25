import { useStyletron } from "styletron-react";
import { useMobile } from "../hooks/useMobile";

type InternalPageContainerProps = {
  children: JSX.Element | JSX.Element[];
};

export const InternalPageContainer = ({ children }: InternalPageContainerProps) => {
  const [css] = useStyletron();
  const [isMobile] = useMobile();

  return (
    <div
      className={css({
        gridColumn: "1 / 4",
        paddingInlineStart: isMobile ? "0" : "32px",
        paddingInlineEnd: isMobile ? "0" : "32px",
      })}
    >
      {children}
    </div>
  );
};
