import {
  BUTTON_KIND,
  BUTTON_SIZE,
  ButtonIcon,
  COLORS,
  type Items,
  MENU_SIZE,
  Menu,
} from "@nilfoundation/ui-kit";
import type { MenuOverrides } from "baseui/menu";
import { useStyletron } from "styletron-react";
import { getRuntimeConfigOrThrow } from "../../runtime-config";
import { ArrowUpRightIcon, QuestionIcon, StatefulPopover, useMobile } from "../../shared";

const menuOverrides: MenuOverrides = {
  List: {
    style: {
      backgroundColor: COLORS.gray800,
    },
  },
};

export const QuestionButton = () => {
  const [css] = useStyletron();
  const [isMobile] = useMobile();
  const { SANDBOX_FEEDBACK_URL, SANDBOX_SUPPORT_URL, SANDBOX_DOCS_URL } = getRuntimeConfigOrThrow();

  const items: Items = [
    {
      label: "Documentation",
      endEnhancer: <ArrowUpRightIcon />,
      href: SANDBOX_DOCS_URL,
    },
    {
      label: "Support",
      endEnhancer: <ArrowUpRightIcon />,
      href: SANDBOX_SUPPORT_URL,
    },
    {
      label: (
        <span
          className={css({
            whiteSpace: "nowrap",
          })}
        >
          {"Share feedback"}
        </span>
      ),
      href: SANDBOX_FEEDBACK_URL,
    },
  ];

  return (
    <StatefulPopover
      popoverMargin={8}
      content={<Menu isDropdown items={items} size={MENU_SIZE.small} overrides={menuOverrides} />}
      placement="bottomRight"
      autoFocus
      triggerType="click"
    >
      <ButtonIcon
        className={css({
          width: isMobile ? "32px" : "48px",
          height: isMobile ? "32px" : "48px",
          flexShrink: 0,
        })}
        icon={<QuestionIcon />}
        kind={BUTTON_KIND.secondary}
        size={isMobile ? BUTTON_SIZE.compact : BUTTON_SIZE.large}
      />
    </StatefulPopover>
  );
};
