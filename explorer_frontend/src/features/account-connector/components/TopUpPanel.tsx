import {
  Button,
  COLORS,
  Input,
  LabelXSmall,
  NOTIFICATION_KIND,
  Notification,
  ParagraphXSmall,
} from "@nilfoundation/ui-kit";
import { FormControl } from "baseui/form-control";
import type { InputOverrides } from "baseui/input";
import { LabelMedium, LabelSmall } from "baseui/typography";
import { useUnit } from "effector-react";
import { useEffect, useState } from "react";
import { useStyletron } from "styletron-react";
import { getRuntimeConfigOrThrow } from "../../runtime-config";
import { Token } from "../../tokens";
import { TokenInput } from "../../tokens";
import { $faucets } from "../../tokens/model";
import { ActiveComponent } from "../ActiveComponent";
import {
  $smartAccount,
  $topUpError,
  $topupInput,
  resetTopUpError,
  setActiveComponent,
  setTopupInput,
  topupSmartAccountTokenFx,
  topupTokenEvent,
} from "../model";
import { validateAmount } from "../validation.ts";
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

const getQuickAmounts = (selectedToken: string): number[] => {
  return selectedToken === Token.NIL ? [0.0001, 0.003, 0.05] : [1, 5, 10];
};

const TopUpPanel = () => {
  const [css] = useStyletron();
  const [topUpError, setTopUpError] = useState("");
  const topUpExecutionError = useUnit($topUpError);
  const [smartAccount, faucets, topupInput, topupInProgress] = useUnit([
    $smartAccount,
    $faucets,
    $topupInput,
    topupSmartAccountTokenFx.pending,
  ]);

  useEffect(() => {
    setTopupInput({ ...topupInput, amount: "" });

    //Reset error when leaving the page
    return () => {
      resetTopUpError();
    };
  }, []);

  const availableTokens = Object.keys(faucets ?? {});
  const quickAmounts = getQuickAmounts(topupInput.token);

  const handleQuickAmountClick = (amount: number) => {
    setTopUpError("");
    setTopupInput({ ...topupInput, amount: amount.toString() });
  };

  const handleButtonPress = () => {
    const error = validateAmount(topupInput.amount, topupInput.token);
    if (error != null) {
      setTopUpError(error);
      return;
    }
    resetTopUpError();
    topupTokenEvent();
  };

  return (
    <div
      className={css({
        display: "flex",
        flexDirection: "column",
      })}
    >
      <BackLink
        title="Back"
        goBackCb={() => {
          setTopupInput({ token: topupInput.token, amount: "" });
          setActiveComponent(ActiveComponent.Main);
        }}
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
          tokens={availableTokens.map((t) => ({
            token: t,
          }))}
          onChange={({ amount, token }) => {
            setTopUpError("");
            setTopupInput({
              token,
              amount: token !== topupInput.token ? "" : amount,
            });
          }}
          value={{
            token: topupInput.token,
            amount: topupInput.amount,
          }}
        />
      </div>
      {/* Display Error Message Below Input */}
      {topUpError && (
        <LabelSmall style={{ color: COLORS.red400, marginBottom: "8px" }}>{topUpError}</LabelSmall>
      )}

      {/* Quick Amount Buttons */}
      <div
        className={css({
          display: "flex",
          gap: "8px",
          flexDirection: "row",
          marginBottom: "12px",
        })}
      >
        {quickAmounts.map((quickAmount) => (
          <div
            key={quickAmount}
            className={css({
              width: "54px",
              height: "32px",
              backgroundColor: COLORS.gray600,
              color: COLORS.gray200,
              borderRadius: "8px",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              transition: "all 0.2s ease",
              cursor: "pointer",
              ":hover": { backgroundColor: COLORS.gray700 },
              ":active": { backgroundColor: COLORS.gray600, transform: "scale(0.98)" },
            })}
            onClick={() => handleQuickAmountClick(quickAmount)}
          >
            <ParagraphXSmall>{quickAmount}</ParagraphXSmall>
          </div>
        ))}
      </div>

      {topupInput.token === Token.NIL && topUpExecutionError === "" && (
        <Notification
          closeable={true}
          kind={NOTIFICATION_KIND.warning}
          hideIcon={true}
          overrides={{
            Body: {
              style: {
                backgroundColor: COLORS.yellow300,
                marginLeft: 0,
                marginRight: 0,
                width: "100%",
              },
            },
          }}
        >
          <LabelSmall>
            The NIL faucet is capped. The amount received may be different than requested
          </LabelSmall>
        </Notification>
      )}

      {topUpExecutionError && (
        <Notification
          closeable
          kind={NOTIFICATION_KIND.negative}
          hideIcon
          overrides={{
            Body: {
              style: {
                backgroundColor: COLORS.red300,
                marginLeft: 0,
                marginRight: 0,
                width: "100%",
              },
            },
          }}
        >
          <LabelSmall>{topUpExecutionError}</LabelSmall>
        </Notification>
      )}
      <Button
        className={css({
          width: "100%",
          marginTop: "8px",
          marginBottom: "16px",
        })}
        onClick={handleButtonPress}
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
      <LabelXSmall
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
      </LabelXSmall>
    </div>
  );
};

export { TopUpPanel };
