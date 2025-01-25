import { Link } from "atomic-router-react";
import { useStyletron } from "styletron-react";
import { explorerRoute } from "../../../routing/routes/explorerRoute";
import logo from "./assets/Logo.svg";
import { styles } from "./styles";

export const Logo = () => {
  const [css] = useStyletron();

  return (
    <Link className={css(styles.logo)} to={explorerRoute}>
      <img src={logo} alt="logo" />
    </Link>
  );
};
