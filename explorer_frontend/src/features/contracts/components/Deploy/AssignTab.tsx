import type { Hex } from "@nilfoundation/niljs";
import { Button, COLORS, FormControl, Input, SPACE } from "@nilfoundation/ui-kit";
import { INPUT_KIND, INPUT_SIZE } from "@nilfoundation/ui-kit";
import type { InputOverrides } from "baseui/input";
import { useUnit } from "effector-react";
import { useStyletron } from "styletron-react";
import { $smartAccount } from "../../../account-connector/model";
import {
  $assignedSmartContractAddress,
  assignSmartContract,
  assignSmartContractFx,
  setAssignedSmartContractAddress,
} from "../../models/base";

export const AssignTab = () => {
  const [smartAccount, pending, assignedAddress] = useUnit([
    $smartAccount,
    assignSmartContractFx.pending,
    $assignedSmartContractAddress,
  ]);
  const [css] = useStyletron();

  return (
    <>
      <div
        className={css({
          flexGrow: 0,
          paddingBottom: SPACE[16],
        })}
      >
        <FormControl label="Address">
          <Input
            kind={INPUT_KIND.secondary}
            size={INPUT_SIZE.small}
            overrides={inputOverrides}
            onChange={(e) => {
              setAssignedSmartContractAddress(e.target.value as Hex);
            }}
            value={assignedAddress && assignedAddress !== "0x" ? assignedAddress : ""}
          />
        </FormControl>
      </div>
      <div>
        <Button
          onClick={() => {
            assignSmartContract();
          }}
          isLoading={pending}
          disabled={pending || !smartAccount || !assignedAddress}
        >
          Assign
        </Button>
      </div>
    </>
  );
};

const inputOverrides: InputOverrides = {
  Root: {
    style: () => ({
      background: COLORS.gray700,
      ":hover": {
        background: COLORS.gray600,
      },
    }),
  },
};
