import {
  BUTTON_KIND,
  BUTTON_SIZE,
  COLORS,
  ChevronDownIcon,
  ChevronUpIcon,
  type Items,
  MENU_SIZE,
  Menu,
} from "@nilfoundation/ui-kit";
import { Button } from "baseui/button";
import type { MenuOverrides } from "baseui/menu";
import { useUnit } from "effector-react";
import { type FC, useState } from "react";
import { useStyletron } from "styletron-react";
import { StatefulPopover, useMobile } from "../../shared";
import { $availableSolidityVersions, $solidityVersion, changeSolidityVersion } from "../model";

type CompilerVersionButtonProps = {
  disabled?: boolean;
};

export const CompilerVersionButton: FC<CompilerVersionButtonProps> = ({ disabled }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [css] = useStyletron();
  const [isMobile] = useMobile();
  const btnOverrides = {
    Root: {
      style: {
        whiteSpace: "nowrap",
        ...(!isMobile ? { paddingLeft: "24px", paddingRight: "24px" } : {}),
      },
    },
  };

  const version = useUnit($solidityVersion);

  const versions = $availableSolidityVersions.getState().map((v) => {
    return { label: v };
  });

  const menuOverrides: MenuOverrides = {
    List: {
      style: {
        backgroundColor: COLORS.gray800,
        width: "110%",
      },
    },
  };

  return (
    <StatefulPopover
      onOpen={() => setIsOpen(true)}
      onClose={() => setIsOpen(false)}
      popoverMargin={8}
      content={({ close }) => (
        <Menu
          onItemSelect={({ item }) => {
            changeSolidityVersion(item.label);
            close();
          }}
          items={versions as Items}
          size={MENU_SIZE.small}
          overrides={menuOverrides}
          renderAll
          isDropdown
        />
      )}
      placement={isMobile ? "bottomRight" : "bottomLeft"}
      autoFocus
      triggerType="click"
    >
      <Button
        kind={BUTTON_KIND.secondary}
        size={isMobile ? BUTTON_SIZE.compact : BUTTON_SIZE.large}
        className={css({
          height: isMobile ? "32px" : "48px",
          flexShrink: 0,
        })}
        overrides={btnOverrides}
        endEnhancer={isOpen ? <ChevronUpIcon /> : <ChevronDownIcon />}
        disabled={disabled}
      >
        Compiler {version.split("+")[0]}
      </Button>
    </StatefulPopover>
  );
};
