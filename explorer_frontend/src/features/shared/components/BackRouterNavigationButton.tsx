import { ArrowUpIcon, BUTTON_KIND, BUTTON_SIZE, ButtonIcon } from "@nilfoundation/ui-kit";
import { mergeOverrides } from "baseui";
import type { ButtonOverrides } from "baseui/button";
import { useUnit } from "effector-react";
import type { FC } from "react";
import { router } from "../../routing";

type BackButtonProps = {
  overrides?: ButtonOverrides;
  disabled?: boolean;
};

export const BackRouterNavigationButton: FC<BackButtonProps> = ({ overrides, disabled }) => {
  const history = useUnit(router.$history);
  const historyEmpty = window.history.length < 2;
  const mergedOverrides = mergeOverrides(
    {
      Root: {
        style: {
          transform: "rotate(-90deg)",
          width: "48px",
          height: "48px",
        },
      },
    },
    overrides,
  );

  return (
    <ButtonIcon
      icon={<ArrowUpIcon $size={"16px"} />}
      kind={BUTTON_KIND.secondary}
      size={BUTTON_SIZE.large}
      onClick={() => history.back()}
      overrides={mergedOverrides}
      disabled={historyEmpty || disabled}
    />
  );
};
