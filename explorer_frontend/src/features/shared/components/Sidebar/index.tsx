import { useStyletron } from "styletron-react";
import { useMobile } from "../../hooks/useMobile";
import { Navigation } from "./Navigation";
import { styles } from "./styles";

export const Sidebar = () => {
  const [css] = useStyletron();

  const [isMobile] = useMobile();

  if (isMobile) return null;

  return (
    <aside className={css(styles.aside)}>
      <Navigation />
    </aside>
  );
};
