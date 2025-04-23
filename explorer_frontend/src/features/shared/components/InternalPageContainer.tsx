import { useStyletron } from "styletron-react";

type InternalPageContainerProps = {
  children: JSX.Element | JSX.Element[];
};

export const InternalPageContainer = ({ children }: InternalPageContainerProps) => {
  const [css] = useStyletron();

  return (
    <div
      className={css({
        gridColumn: "1 / 4",
        // paddingLeft: isMobile ? "0" : "32px",
        // paddingRight: isMobile ? "0" : "32px",
      })}
    >
      {children}
    </div>
  );
};
