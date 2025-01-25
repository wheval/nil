import {
  Popover as PopoverBase,
  type PopoverProps,
  StatefulPopover as StatefulPopoverBase,
  type StatefulPopoverProps,
} from "baseui/popover";
import type { FC } from "react";

const StatefulPopover: FC<StatefulPopoverProps> = ({ ...props }) => {
  return <StatefulPopoverBase {...props} dismissOnEsc />;
};

const Popover: FC<PopoverProps> = ({ ...props }) => {
  return <PopoverBase {...props} />;
};

export { StatefulPopover, Popover };
