import { useUnit } from "effector-react";
import { useSwipeable } from "react-swipeable";
import { Code } from "../../features/code/Code";
import { ContractsContainer } from "../../features/contracts";
import { Logs } from "../../features/logs/components/Logs";
import { $activeComponent, LayoutComponent, setActiveComponent } from "./model";

const featureMap = new Map();
featureMap.set(LayoutComponent.Code, Code);
featureMap.set(LayoutComponent.Logs, Logs);
featureMap.set(LayoutComponent.Contracts, ContractsContainer);

const SandboxMobileLayout = () => {
  const activeComponent = useUnit($activeComponent);
  const Component = activeComponent ? featureMap.get(activeComponent) : null;
  const handlers = useSwipeable({
    onSwipedLeft: () => setActiveComponent(LayoutComponent.Code),
    onSwipedRight: () => setActiveComponent(LayoutComponent.Code),
  });

  return (
    <div {...handlers}>
      <Component />
    </div>
  );
};

export { SandboxMobileLayout };
