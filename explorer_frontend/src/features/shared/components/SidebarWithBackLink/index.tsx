import { useStyletron } from "styletron-react";
import { useMobile } from "../../hooks/useMobile";
import { BackRouterNavigationButton } from "../BackRouterNavigationButton";
import { styles } from "./styles";

export const SidebarWithBackLink = () => {
  const [css] = useStyletron();

  const [isMobile] = useMobile();

  return (
    <aside className={css(isMobile ? styles.mobileAside : styles.aside)}>
      <BackRouterNavigationButton
        overrides={{
          Root: {
            style: {
              marginLeft: "32px",
            },
          },
        }}
      />
    </aside>
  );
};
