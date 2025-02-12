import { persist } from "effector-storage/session";
import { $activeComponent, setActiveComponent } from "./model";

$activeComponent.on(setActiveComponent, (_, payload) => payload);

persist({
  store: $activeComponent,
  key: "activeComponentPlayground",
});
