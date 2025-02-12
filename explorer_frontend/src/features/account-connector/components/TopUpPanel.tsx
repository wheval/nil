import { Button, COLORS, Input } from "@nilfoundation/ui-kit";
import { FormControl } from "baseui/form-control";
import type { InputOverrides } from "baseui/input";
import { LabelMedium, LabelSmall } from "baseui/typography";
import { useUnit } from "effector-react";
import { useStyletron } from "styletron-react";
import { getRuntimeConfigOrThrow } from "../../runtime-config";
import { TokenInput } from "../../tokens";
import { $faucets } from "../../tokens/model";
import { ActiveComponent } from "../ActiveComponent";
import {
  $smartAccount,
  $topupInput,
  setActiveComponent,
  setTopupInput,
  topupSmartAccountTokenFx,
  topupTokenEvent,
} from "../model";
import { BackLink } from "./BackLink";

const inputOverrides: InputOverrides = {
  Root: {
    style: {
      backgroundColor: COLORS.gray700,
      ":hover": {
        backgroundColor: COLORS.gray600,
      },
    },
  },
};

const TopUpPanel = () => {
  const [css] = useStyletron();
  const [smartAccount, faucets, topupInput, topupInProgress] = useUnit([
    $smartAccount,
    $faucets,
    $topupInput,
    topupSmartAccountTokenFx.pending,
  ]);

  // currently faucet returns mzk so we need to pretend like it is nil token
  const availiableTokens = Object.keys(faucets ?? {});

  return (
    <div
      className={css({
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        paddingRight: "24px",
      })}
    >
      <BackLink
        title="Back"
        goBackCb={() => setActiveComponent(ActiveComponent.Main)}
        disabled={topupInProgress}
      />
      <div
        className={css({
          width: "100%",
          marginTop: "8px",
        })}
      >
        <FormControl label={<LabelMedium>To</LabelMedium>}>
          <Input readOnly placeholder={smartAccount?.address ?? ""} overrides={inputOverrides} />
        </FormControl>
      </div>
      <div
        className={css({
          width: "100%",
        })}
      >
        <TokenInput
          label="Amount"
          tokens={availiableTokens.map((t) => ({
            token: t,
          }))}
          onChange={({ amount, token }) => {
            setTopupInput({
              token,
              amount,
            });
          }}
          value={{
            token: topupInput.token,
            amount: topupInput.amount,
          }}
        />
      </div>
      <Button
        className={css({
          width: "100%",
          marginTop: "8px",
          marginBottom: "16px",
        })}
        onClick={() => topupTokenEvent()}
        disabled={topupInProgress || topupInput.amount === ""}
        isLoading={topupInProgress}
        overrides={{
          Root: {
            style: ({ $disabled }) => ({
              height: "48px",
              backgroundColor: $disabled ? `${COLORS.gray400}!important` : "",
            }),
          },
        }}
      >
        Top up
      </Button>
      <LabelSmall
        color={COLORS.gray200}
        className={css({
          textAlign: "center",
          display: "inline-block",
          fontSize: "14px",
        })}
      >
        <a
          href={getRuntimeConfigOrThrow().PLAYGROUND_MULTI_TOKEN_URL}
          target="_blank"
          rel="noreferrer"
          className={css({
            textDecoration: "underline",
          })}
        >
          Learn
        </a>{" "}
        how to handle tokens and multi-tokens in your environment.
      </LabelSmall>
    </div>
  );
};

export { TopUpPanel };
