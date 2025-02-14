import { COLORS } from "@nilfoundation/ui-kit";
import { useUnit } from "effector-react";
import { expandProperty } from "inline-style-expand-shorthand";
import { useSwipeable } from "react-swipeable";
import { useStyletron } from "styletron-react";
import { ActiveComponent } from "../ActiveComponent";
import { $activeComponent, setActiveComponent } from "../model";
import EndpointScreen from "./EndpointScreen.tsx";
import { MainScreen } from "./MainScreen";
import { TopUpPanel } from "./TopUpPanel";

const featureMap = new Map();
featureMap.set(ActiveComponent.Endpoint, EndpointScreen);
featureMap.set(ActiveComponent.Main, MainScreen);
featureMap.set(ActiveComponent.Topup, TopUpPanel);

const AccountContainer = () => {
  const activeComponent = useUnit($activeComponent);
  const Component = activeComponent ? featureMap.get(activeComponent) : null;
  const [css] = useStyletron();
  const handlers = useSwipeable({
    onSwipedLeft: () => setActiveComponent(ActiveComponent.Endpoint),
    onSwipedRight: () => setActiveComponent(ActiveComponent.Endpoint),
  });

  return (
    <div
      {...handlers}
      className={css({
        ...expandProperty("padding", "24px"),
        ...expandProperty("borderRadius", "16px"),
        width: "100%",
        maxWidth: "420px",
        backgroundColor: COLORS.gray800,
        "@media (min-width: 421px)": {
          width: "420px",
          margin: "0 auto",
        },
      })}
    >
      <Component />
    </div>
  );
};

export { AccountContainer };
