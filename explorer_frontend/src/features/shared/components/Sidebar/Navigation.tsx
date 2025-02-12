import {
  COLORS,
  ChartIcon,
  CodeIcon,
  type LinkComponentRenderFunction,
  MENU_SIZE,
  Menu,
} from "@nilfoundation/ui-kit";
import { Link, useRouter } from "atomic-router-react";
import type { Items, MenuOverrides } from "baseui/menu";
import { useUnit } from "effector-react";
import { playgroundRoute } from "../../../routing";
import { explorerRoute } from "../../../routing/routes/explorerRoute";
import { BackRouterNavigationButton } from "../BackRouterNavigationButton";

const menuOverrides: MenuOverrides = {
  List: {
    style: {
      boxShadow: "none",
      maxWidth: "171px",
    },
  },
};

export const Navigation = () => {
  const router = useRouter();

  const [activeRoute] = useUnit(router.$activeRoutes);
  const isMainPage = activeRoute === explorerRoute;
  const isPlayground = activeRoute === playgroundRoute;

  const items: Items = [
    {
      label: "Explorer",
      startEnhancer: <ChartIcon />,
      isHighlighted: isMainPage,
      linkComponent: (({ children, className }) => (
        <Link to={explorerRoute} className={className}>
          {children}
        </Link>
      )) as LinkComponentRenderFunction,
    },
    {
      label: "Playground",
      startEnhancer: <CodeIcon $color={COLORS.gray100} />,
      isHighlighted: isPlayground,
      linkComponent: (({ children, className }) => (
        <Link to={playgroundRoute} className={className}>
          {children}
        </Link>
      )) as LinkComponentRenderFunction,
    },
    {
      label: "Diagnostic",
      startEnhancer: <ChartIcon />,
      disabled: true,
    },
  ];

  if (!isMainPage) {
    return <BackRouterNavigationButton />;
  }

  return <Menu items={items} size={MENU_SIZE.small} overrides={menuOverrides} />;
};
