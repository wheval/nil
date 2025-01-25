import { COLORS } from "@nilfoundation/ui-kit";
import { useUnit } from "effector-react";
import { expandProperty } from "inline-style-expand-shorthand";
import { useSwipeable } from "react-swipeable";
import { useStyletron } from "styletron-react";
import { ActiveComponent } from "../ActiveComponent";
import { $activeComponent, setActiveComponent } from "../model";
import { MainScreen } from "./MainScreen";
import { TopUpPanel } from "./TopUpPanel";

const featureMap = new Map();
featureMap.set(ActiveComponent.Main, MainScreen);
featureMap.set(ActiveComponent.Topup, TopUpPanel);

const AccountContainer = () => {
  const activeComponent = useUnit($activeComponent);
  const Component = activeComponent ? featureMap.get(activeComponent) : null;
  const [css] = useStyletron();
  const handlers = useSwipeable({
    onSwipedLeft: () => setActiveComponent(ActiveComponent.Main),
    onSwipedRight: () => setActiveComponent(ActiveComponent.Main),
  });

  return (
    <div
      {...handlers}
      className={css({
        ...expandProperty("padding", "24px 0 24px 24px"),
        ...expandProperty("borderRadius", "16px"),
        width: "420px",
        backgroundColor: COLORS.gray800,
      })}
    >
      <Component />
    </div>
  );
};

export { AccountContainer };
